import { createFileRoute } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';
import { getPost } from '@/lib/blog';
import ReactMarkdown, { type Components } from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github-dark.css';

export const Route = createFileRoute('/blog/$slug')({
  loader: ({ params }) => {
    const post = getPost(params.slug);
    if (!post) throw new Error('Post not found');
    return { post };
  },
  head: ({ loaderData }) => {
    if (!loaderData?.post) {
      return { title: 'Blog Post' };
    }

    return {
      title: loaderData.post.title,
      meta: [
        { name: 'description', content: loaderData.post.description },
        { property: 'og:title', content: loaderData.post.title },
        { property: 'og:description', content: loaderData.post.description },
        { property: 'og:image', content: loaderData.post.image },
        { property: 'og:type', content: 'article' },
      ],
    };
  },
  component: BlogPost,
});

function BlogPost() {
  const { post } = Route.useLoaderData();

  const components: Components = {
    a: ({ href, children }) => {
      const isExternal =
        href && (href.startsWith('http') || href.startsWith('https'));

      return (
        <a
          href={href}
          className="text-blue-400 hover:text-blue-500 underline"
          target={isExternal ? '_blank' : undefined}
          rel={isExternal ? 'noopener noreferrer' : undefined}
        >
          {children}
        </a>
      );
    },
  };

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
        <h1 className="text-4xl font-bold">{post.title}</h1>
        <time className="text-sm text-gray-600 mb-2 block">
          {formatDate(post.date)}
        </time>
      </header>

      <div className="prose prose-slate dark:prose-invert max-w-none">
        <ReactMarkdown
          rehypePlugins={[rehypeHighlight]}
          components={components}
        >
          {post.content}
        </ReactMarkdown>
      </div>
    </article>
  );
}
