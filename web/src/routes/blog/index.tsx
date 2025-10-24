import { createFileRoute } from '@tanstack/react-router';
import { getAllPosts } from '@/lib/blog';
import { BlogCard } from '@/components/BlogCard';

export const Route = createFileRoute('/blog/')({
  loader: () => {
    const posts = getAllPosts();
    return { posts };
  },
  component: BlogPage,
});

function BlogPage() {
  const { posts } = Route.useLoaderData();

  return (
    <div className="max-w-[1400px] mx-auto px-8 py-12">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4">Welcome to my Blog!</h1>
        <p className="text-lg text-gray-600 max-w-2xl mx-auto">
          Here I will discuss topics that interest me, projects I am working on,
          and anything I think deserves a post!
        </p>
      </div>

      <div className="grid grid-cols-3 gap-8">
        {posts.map((post) => (
          <BlogCard
            key={post.slug}
            slug={post.slug}
            title={post.title}
            description={post.description}
            date={post.date}
            image={post.image}
          />
        ))}
      </div>

      {posts.length === 0 && (
        <div className="text-center py-16">
          <p className="text-gray-500">No blog posts yet. Check back soon!</p>
        </div>
      )}
    </div>
  );
}
