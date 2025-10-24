import { AboutCard } from '@/components/about/AboutCard.tsx';

export function JourneyTimelineCard() {
  const timelineItems = [
    {
      year: '2017-2024',
      title: 'Sales Associate & IT Assistant',
      company: 'Ace Hardware',
    },
    { year: '2021', title: 'Started CS Degree', company: 'University' },
    { year: '2023', title: 'Discovered Authelia', company: 'Open Source' },
    { year: '2024', title: 'Became Maintainer', company: 'Authelia' },
    { year: '2024', title: 'Software Development Intern', company: 'Current' },
  ];

  return (
    <AboutCard
      title="Professional Journey"
      description="My path through technology"
    >
      <div className="relative">
        {timelineItems.map((item, index) => (
          <div key={index} className="flex items-start mb-4 last:mb-0">
            <div className="flex-shrink-0 w-20 text-sm text-muted-foreground">
              {item.year}
            </div>
            <div className="flex-shrink-0 w-3 h-3 bg-primary rounded-full mt-1.5 mx-4"></div>
            <div className="flex-1">
              <h4 className="font-medium">{item.title}</h4>
              <p className="text-sm text-muted-foreground">{item.company}</p>
            </div>
          </div>
        ))}
      </div>
    </AboutCard>
  );
}
