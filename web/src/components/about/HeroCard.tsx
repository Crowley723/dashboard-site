import { AboutCard } from '@/components/about/AboutCard.tsx';

export function HeroCard() {
  return (
    <AboutCard title="Brynn Crowley" description="He/Him">
      <div className="space-y-2">
        <p>
          Brynn Crowley is a dedicated computer science student pursuing a
          Bachelor of Science. Passionate about technology and its practical
          applications, Brynn combines academic learning with hands-on
          experience to develop a well-rounded skill set in the field of
          Computer Science and Cybersecurity.
        </p>
        <p>
          Recently, Brynn has been focusing on building and maintaining a
          comprehensive homelab environment. This project encompasses various
          technologies including Docker containerization, Kubernetes container
          orchestration, network management using OPNsense, and distributed
          application architecture. Additionally, Brynn is a maintainer for the
          open-source authentication project Authelia and other minor projects.
        </p>
      </div>
    </AboutCard>
  );
}
