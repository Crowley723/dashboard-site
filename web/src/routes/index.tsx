import { createFileRoute } from '@tanstack/react-router';
import { PodUptimeCards } from '@/components/PodUptimeCards.tsx';
import { NodeStatusCard } from '@/components/NodeStatusCard.tsx';
import { PodsPerNamespaceCard } from '@/components/PodsPerNamespaceCard.tsx';
import { ClusterAvgCPUCard } from '@/components/ClusterAvgCPUCard.tsx';
import { RecentPodRestartsCard } from '@/components/RecentPodRestartsCard.tsx';
import { TraefikAvgRequestsCard } from '@/components/TraefikAvgRequestsCard.tsx';

export const Route = createFileRoute('/')({
  component: Index,
});

function Index() {
  return (
    <div className="grid grid-cols-4 gap-4 justify-items-center my-4">
      <div className="row-span-2 p-2">
        <PodUptimeCards />
      </div>
      <div className="p-2">
        <NodeStatusCard />
      </div>
      <div className="p-2">
        <PodsPerNamespaceCard />
      </div>
      <div className="p-2">
        <RecentPodRestartsCard />
      </div>
      <div className="p-2">
        <ClusterAvgCPUCard />
      </div>
      <div className="p-2">
        <TraefikAvgRequestsCard />
      </div>
    </div>
  );
}
