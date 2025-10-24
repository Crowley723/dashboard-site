import { Link } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';

interface BlogCardProps {
  slug: string;
  title: string;
  description?: string;
  date: string | Date;
  image?: string;
  readingTime?: number;
}

export function BlogCard({
  slug,
  title,
  description,
  date,
  image,
  readingTime,
}: BlogCardProps) {
  const formatDate = (dateValue: string | Date) => {
    if (!dateValue) return '';

    if (dateValue instanceof Date) {
      return format(dateValue, 'MMMM d, yyyy');
    }

    try {
      return format(parseISO(dateValue), 'MMMM d, yyyy');
    } catch (error) {
      return String(dateValue);
    }
  };

  return (
    <Link to="/blog/$slug" params={{ slug }} className="block group">
      <article className="h-[360px] border rounded-lg overflow-hidden transition-transform hover:scale-105">
        <div className="h-40 bg-gray-200 overflow-hidden">
          {image ? (
            <img
              src={image}
              alt={title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-gray-400">
              <span className="text-5xl">📝</span>
            </div>
          )}
        </div>

        <div className="p-5 h-[200px] flex flex-col">
          <time className="text-xs text-gray-500 mb-2 block uppercase tracking-wide">
            {formatDate(date)}
            {readingTime ? ` • ${readingTime} min read` : ''}
          </time>
          <h2 className="text-lg font-semibold mb-3 line-clamp-2 ">{title}</h2>

          {description && (
            <>
              <p className="text-sm text-gray-600 line-clamp-4 flex-1 leading-relaxed">
                {description}
              </p>
              <span className="group-hover:text-blue-600 text-xs font-medium">
                Read more →
              </span>
            </>
          )}
        </div>
      </article>
    </Link>
  );
}
