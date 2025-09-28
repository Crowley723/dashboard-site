import { createFileRoute } from '@tanstack/react-router';
import { PodUptimeCards } from '@/components/PodUptimeCards.tsx';
import { NodeStatusCard } from '@/components/NodeStatusCard.tsx';
import { PodsPerNamespaceCard } from '@/components/PodsPerNamespaceCard.tsx';
import { ClusterAvgCPUHourCard } from '@/components/ClusterAvgCPUHourCard.tsx';
import { TraefikAvgReqPerSecHourCard } from '@/components/TraefikAvgReqPerSecHourCard.tsx';
import { TraefikTotalRequestsHourCard } from '@/components/TraefikTotalRequestsHourCard.tsx';
import { TraefikAvgReqPerSecWeekCard } from '@/components/TraefikAvgReqPerSecWeekCard.tsx';
import { TraefikTotalRequestsWeekCard } from '@/components/TraefikTotalRequestsWeekCard.tsx';
import { ClusterAvgCPUWeekCard } from '@/components/ClusterAvgCPUWeekCard.tsx';

export const Route = createFileRoute('/')({
  component: Index,
});

function Index() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 auto-rows-72 gap-12 p-4">
      <div className="col-span-1 row-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <PodUptimeCards />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <NodeStatusCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <PodsPerNamespaceCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1"></div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <ClusterAvgCPUHourCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <ClusterAvgCPUWeekCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <TraefikAvgReqPerSecHourCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <TraefikAvgReqPerSecWeekCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <TraefikTotalRequestsHourCard />
      </div>
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <TraefikTotalRequestsWeekCard />
      </div>
    </div>
  );
}
