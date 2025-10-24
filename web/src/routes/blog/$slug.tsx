import { createFileRoute } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';
import { getPost } from '@/lib/blog';
import ReactMarkdown from 'react-markdown';

export const Route = createFileRoute('/blog/$slug')({
  loader: ({ params }) => {
    const post = getPost(params.slug);
    if (!post) throw new Error('Post not found');
    return { post };
  },
  component: BlogPost,
});

function BlogPost() {
  const { post } = Route.useLoaderData();

  const formatDate = (dateValue: string | Date) => {
    if (!dateValue) return '';

    if (dateValue instanceof Date) {
      return format(dateValue, 'LLLL d, yyyy');
    }

    try {
      return format(parseISO(dateValue), 'LLLL d, yyyy');
    } catch (error) {
      return String(dateValue);
    }
  };

  return (
    <article className="max-w-3xl mx-auto py-12 px-8">
      <header className="mb-8 text-center border-b pb-8">
        <time className="text-sm text-gray-600 mb-2 block">
          {formatDate(post.date)}
        </time>
        <h1 className="text-4xl font-bold">{post.title}</h1>
      </header>

      {/* Just add prose class - that's it! */}
      <div className="prose prose-lg max-w-none">
        <ReactMarkdown>{post.content}</ReactMarkdown>
      </div>
    </article>
  );
}
