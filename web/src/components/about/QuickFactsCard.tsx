import { AboutCard } from '@/components/about/AboutCard.tsx';

export function QuickFactsCard() {
  return (
    <AboutCard title="Quick Facts" description="At a glance">
      <div className="space-y-2">
        <p>Location: San Francisco, Bay Area, California, USA</p>
        <p>Graduating: Spring 2026</p>
        {/*<p>Currently: Software Development Intern</p>*/}
        <p>Currently: IT Assistant</p>
      </div>
    </AboutCard>
  );
}
