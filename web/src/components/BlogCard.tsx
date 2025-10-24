import { Link } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';

interface BlogCardProps {
  slug: string;
  title: string;
  description?: string;
  date: string | Date;
  image?: string;
}

export function BlogCard({
  slug,
  title,
  description,
  date,
  image,
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
      <article className="h-[320px] border rounded-lg overflow-hidden transition-transform hover:scale-105">
        {/* Image */}
        <div className="h-48 bg-gray-200 overflow-hidden">
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

        {/* Content */}
        <div className="p-4 h-[128px] flex flex-col">
          <time className="text-xs text-gray-600 mb-2">{formatDate(date)}</time>

          <h2 className="text-lg font-semibold mb-2 line-clamp-2 group-hover:text-blue-600">
            {title}
          </h2>

          {description && (
            <p className="text-sm text-gray-600 line-clamp-2 flex-1">
              {description}
            </p>
          )}
        </div>
      </article>
    </Link>
  );
}
