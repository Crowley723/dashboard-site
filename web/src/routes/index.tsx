import { createFileRoute, useRouter } from '@tanstack/react-router';
import { PodUptimeCards } from '@/components/PodUptimeCards.tsx';
import { NodeStatusCard } from '@/components/NodeStatusCard.tsx';
import { PodsPerNamespaceCard } from '@/components/PodsPerNamespaceCard.tsx';
import { ClusterAvgCPUHourCard } from '@/components/ClusterAvgCPUHourCard.tsx';
import { TraefikAvgReqPerSecHourCard } from '@/components/TraefikAvgReqPerSecHourCard.tsx';
import { TraefikTotalRequestsHourCard } from '@/components/TraefikTotalRequestsHourCard.tsx';
import { TraefikAvgReqPerSecWeekCard } from '@/components/TraefikAvgReqPerSecWeekCard.tsx';
import { TraefikTotalRequestsWeekCard } from '@/components/TraefikTotalRequestsWeekCard.tsx';
import { ClusterAvgCPUWeekCard } from '@/components/ClusterAvgCPUWeekCard.tsx';
import { useAuth } from '@/hooks/useAuth.tsx';
import { useEffect, useState } from 'react';
import { LoginDialog } from '@/components/LoginDialog.tsx';

export const Route = createFileRoute('/')({
  component: Index,
  validateSearch: (search: Record<string, unknown>) => {
    return {
      showLogin: search.showLogin === true || undefined,
      message: (search.message as string) || undefined,
      rd: (search.rd as string) || undefined,
      error: search.error || undefined,
    };
  },
});

function Index() {
  const { isLoading, login, isLoggingIn } = useAuth();
  const { showLogin, error, message, rd } = Route.useSearch();
  const [showLoginModal, setShowLoginModal] = useState(false);
  const router = useRouter();

  useEffect(() => {
    if (showLogin) {
      setShowLoginModal(true);
    }
  }, [showLogin]);

  const handleCloseModal = () => {
    setShowLoginModal(false);
    router.navigate({ to: '/', search: {} as any });
  };

  const handleLogin = () => {
    login(rd, {
      onSuccess: () => {
        setShowLoginModal(false);
        router.navigate({ to: '/', search: {} as any });
        if (rd) {
          router.navigate({ to: rd });
        }
      },
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        Loading...
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 auto-rows-96 gap-4 p-4">
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

      {showLogin && (
        <LoginDialog
          login={handleLogin}
          isLoggingIn={isLoggingIn}
          open={showLoginModal}
          onOpenChange={(open) => {
            setShowLoginModal(open);
            if (!open) handleCloseModal();
          }}
          message={message}
          error={error as boolean}
        >
          <div></div>
        </LoginDialog>
      )}
    </>
  );
}
