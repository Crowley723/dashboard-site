import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface UserDisplayProps {
  displayName: string;
  username: string;
  sub: string;
  iss: string;
  className?: string;
}

export function UserDisplay({
  displayName,
  username,
  sub,
  iss,
  className,
}: UserDisplayProps) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className={`cursor-help decoration-muted-foreground/50 underline decoration-dotted underline-offset-4 ${className || ''}`}
          >
            {displayName || username}
          </span>
        </TooltipTrigger>
        <TooltipContent>
          <div className="space-y-1 text-sm">
            <div className="font-medium">{username}</div>
            <div className="text-muted-foreground text-xs">
              {sub}@{iss}
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
