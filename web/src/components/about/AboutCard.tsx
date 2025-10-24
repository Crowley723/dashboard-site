import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card.tsx';
import type { ReactNode } from 'react';

interface AboutCardProps {
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}

export function AboutCard({
  title,
  description,
  children,
  className,
}: AboutCardProps) {
  return (
    <Card className={`h-full ${className || ''}`}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
