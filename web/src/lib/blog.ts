import matter from 'gray-matter';

export interface Post {
  slug: string;
  title: string;
  date: string;
  content: string;
  url: string;
  description?: string;
  image?: string;
  readingTime?: number;
}

const postModules = import.meta.glob<string>('/content/blog/*.md', {
  query: '?raw',
  import: 'default',
  eager: true,
});

function calculateReadingTime(content: string): number {
  const wordsPerMinute = 200;
  if (content == null) {
    return 0;
  }
  const words = content.trim().split(/\s+/).length;
  return Math.ceil(words / wordsPerMinute);
}

export function getAllPosts(): Post[] {
  const allPosts = Object.entries(postModules).map(([filepath, content]) => {
    const { data, content: markdown } = matter(content);
    const slug = filepath.split('/').pop()?.replace('.md', '') || '';
    console.log(markdown);
    const readingTime = calculateReadingTime(markdown);

    return {
      slug,
      title: data.title as string,
      date: data.date as string,
      content: markdown,
      url: `/blog/${slug}`,
      description: data.description as string | undefined,
      image: data.image as string | undefined,
      readingTime: readingTime,
    };
  });

  return allPosts.sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
  );
}

export function getPost(slug: string): Post | undefined {
  return getAllPosts().find((post) => post.slug === slug);
}
