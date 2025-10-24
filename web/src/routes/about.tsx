import { createFileRoute } from '@tanstack/react-router';
import { HeroCard } from '@/components/about/HeroCard.tsx';
import { JourneyTimelineCard } from '@/components/about/JourneyTimelineCard.tsx';
import { TechnicalSkillsCard } from '@/components/about/TechnicalSkillsCard.tsx';
import { CurrentProjectsCard } from '@/components/about/CurrentProjectsCard.tsx';
import { OpenSourceCard } from '@/components/about/OpenSourceCard.tsx';
import { HomelabStatsCard } from '@/components/about/HomelabStatsCard.tsx';
import { ContactLinksCard } from '@/components/about/ContactLinksCard.tsx';
import { EducationCard } from '@/components/about/EducationCard.tsx';
import { QuickFactsCard } from '@/components/about/QuickFactsCard.tsx';

export const Route = createFileRoute('/about')({
  component: About,
});

function About() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 auto-rows-72 gap-12 p-4">
      {/* Hero section - spans 2 columns on larger screens */}
      <div className="col-span-1 sm:col-span-2 lg:col-span-2 lg:row-span-2">
        <HeroCard />
      </div>

      {/* Quick facts */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <QuickFactsCard />
      </div>

      {/* Contact links */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <ContactLinksCard />
      </div>

      {/* Journey timeline - wide span */}
      <div className="col-span-1 sm:col-span-2 lg:col-span-3 lg:row-span-1">
        <JourneyTimelineCard />
      </div>

      {/* Empty space for balance */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1"></div>

      {/* Technical skills */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <TechnicalSkillsCard />
      </div>

      {/* Current projects */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <CurrentProjectsCard />
      </div>

      {/* Open source contributions */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <OpenSourceCard />
      </div>

      {/* Education */}
      <div className="col-span-1 sm:col-span-1 lg:col-span-1 lg:row-span-1">
        <EducationCard />
      </div>

      {/* Homelab stats - spans 2 columns */}
      <div className="col-span-1 sm:col-span-2 lg:col-span-2 lg:row-span-1">
        <HomelabStatsCard />
      </div>
    </div>
  );
}
