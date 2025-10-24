import matter from 'gray-matter';

export interface Post {
  slug: string;
  title: string;
  date: string;
  content: string;
  url: string;
  description?: string;
  image?: string;
}

const postModules = import.meta.glob<string>('/content/blog/*.md', {
  query: '?raw',
  import: 'default',
  eager: true,
});

export function getAllPosts(): Post[] {
  const allPosts = Object.entries(postModules).map(([filepath, content]) => {
    const { data, content: markdown } = matter(content);
    const slug = filepath.split('/').pop()?.replace('.md', '') || '';

    return {
      slug,
      title: data.title as string,
      date: data.date as string,
      content: markdown,
      url: `/blog/${slug}`,
      description: data.description as string | undefined,
      image: data.image as string | undefined,
    };
  });

  return allPosts.sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
  );
}

export function getPost(slug: string): Post | undefined {
  return getAllPosts().find((post) => post.slug === slug);
}
