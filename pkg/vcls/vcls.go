package vcls

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

type serviceImpl struct {
	store     Store
	cfsClient cfs.Client
	repo      mysql.VMRepository
	witness   witness.Client
	fdmAgent  fdm.Agent
	cache        cache.Cache[string, bool] // simple cache to debounce Refresh
	cacheManager cache.CacheManager        // optional; nil if not wired
	metrics      metrics.Metrics
	log          logger.Logger
}

var _ Service = &serviceImpl{}

// NewServiceWithCacheManager is identical to NewService but additionally
// wires a CacheManager so that Refresh can track protected VMs for
// background cache synchronization. Pass nil cacheManager to skip.
func NewServiceWithCacheManager(
	store Store, cfsClient cfs.Client, repo mysql.VMRepository,
	witness witness.Client, fdmAgent fdm.Agent,
	cache cache.Cache[string, bool], cacheManager cache.CacheManager,
	m metrics.Metrics, log logger.Logger,
) Service {
	s := NewService(store, cfsClient, repo, witness, fdmAgent, cache, m, log).(*serviceImpl)
	s.cacheManager = cacheManager
	return s
}

func NewService(store Store, cfsClient cfs.Client, repo mysql.VMRepository, witness witness.Client, fdmAgent fdm.Agent, cache cache.Cache[string, bool], m metrics.Metrics, log logger.Logger) Service {
	return &serviceImpl{
		store:     store,
		cfsClient: cfsClient,
		repo:      repo,
		witness:   witness,
		fdmAgent:  fdmAgent,
		cache:     cache,
		metrics:   m,
		log:       log,
	}
}

func (s *serviceImpl) Refresh(ctx context.Context, clusterID string) error {
	// Basic debouncing via cache
	cacheKey := fmt.Sprintf("refresh_%s", clusterID)
	if s.cache != nil {
		if _, ok := s.cache.Get(cacheKey); ok {
			return nil
		}
	}

	// 1. Fetch raw VMs from CFS
	cfsVMs, err := s.cfsClient.ListVMs(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to list VMs from CFS: %w", err)
	}

	// 2. Fetch protected VMs from Repo to establish map
	protectedVMs, err := s.repo.ListProtected(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to list protected VMs from DB: %w", err)
	}

	protectedMap := make(map[string]bool)
	for _, vm := range protectedVMs {
		protectedMap[vm.VMID] = true
	}

	// 3. Batch check witnesses
	var allVMIDs []string
	for _, vm := range cfsVMs {
		allVMIDs = append(allVMIDs, vm.ID)
	}

	var witnessMap map[string]*witness.WitnessState
	if s.witness != nil {
		witnessMap, _ = s.witness.CheckBatch(ctx, allVMIDs)
	}

	// 4. Get FDM ClusterView to establish host health map
	var cv fdm.ClusterView
	if s.fdmAgent != nil {
		cv = s.fdmAgent.ClusterView()
	}

	now := time.Now()
	var eligibleCount int

	// 5 & 6. Aggregate
	for _, cfsVM := range cfsVMs {
		vm := &VM{
			ID:               cfsVM.ID,
			ClusterID:        cfsVM.ClusterID,
			CurrentHost:      cfsVM.HostID,
			PowerState:       PowerStatus(cfsVM.PowerState),
			Protected:        protectedMap[cfsVM.ID],
			WitnessAvailable: true, // default
			HostHealthy:      true, // default
			LastSyncAt:       now,
		}

		if witnessMap != nil {
			if ws, ok := witnessMap[cfsVM.ID]; ok {
				vm.WitnessAvailable = ws.Available
			}
		}

		if s.fdmAgent != nil {
			state, ok := cv.Nodes[cfsVM.HostID]
			if ok && state != fdm.NodeStateAlive {
				vm.HostHealthy = false
			}
		}

		// FDM level checks based on cluster view state
		degradationLevel := cv.DegradationLevel

		// EligibleForHA = Protected && !HostHealthy && PowerState==Running
		// We can add logic to restrict HA based on degradation level here
		vm.EligibleForHA = vm.Protected && !vm.HostHealthy && vm.PowerState == PowerRunning
		if degradationLevel == fdm.OldDegradationAll {
			vm.EligibleForHA = false // cluster totally failed
		}

		if vm.EligibleForHA {
			eligibleCount++
		}

		// 7. Write to store
		s.store.Put(vm)

		if s.cacheManager != nil {
			if vm.Protected {
				s.cacheManager.TrackVM(vm.ID)
			} else {
				s.cacheManager.UntrackVM(vm.ID)
			}
		}
	}

	if s.metrics != nil {
		s.metrics.SetProtectedVMCount(clusterID, float64(len(protectedMap)))
	}

	// update cache TTL (e.g. 5 seconds debounce)
	if s.cache != nil {
		s.cache.Set(cacheKey, true, 5*time.Second)
	}

	if s.log != nil {
		s.log.Info("VCLS Refreshed", "clusterID", clusterID, "vms", len(cfsVMs), "eligible", eligibleCount)
	}
	return nil
}

func (s *serviceImpl) GetVM(ctx context.Context, vmID string) (*VM, error) {
	vm, ok := s.store.Get(vmID)
	if !ok {
		return nil, fmt.Errorf("vm not found in view")
	}
	return vm, nil
}

func (s *serviceImpl) ListProtected(ctx context.Context, clusterID string) ([]*VM, error) {
	var result []*VM
	for _, vm := range s.store.List(clusterID) {
		if vm.Protected {
			result = append(result, vm)
		}
	}
	return result, nil
}

func (s *serviceImpl) ListEligible(ctx context.Context, clusterID string) ([]*VM, error) {
	return s.store.ListEligible(clusterID), nil
}

