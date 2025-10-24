/// <reference types="vite/client" />

declare module 'virtual:blog-posts' {
  export const posts: Array<{
    slug: string;
    title: string;
    date: string;
    content: string;
    url: string;
  }>;
}
