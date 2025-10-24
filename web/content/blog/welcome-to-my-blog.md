---
title: Welcome to my blog!
date: 2025-10-23
description: An introduction to my blog covering open source contributions, homelab experiments, and development learnings
image: https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=800&h=450&fit=crop
---

After suggestions from some of my friends, I have decided to add a blog to my website.

A little about myself: I am one semester away from completing my BS in Computer Science. I spend a lot of my time working on my homelab, adding security features, breaking working applications, generally making a nuisance of myself, contributing to the open-source single sign on provider [Authelia], or working on personal projects. 

The goal with this blog is to give me a place to formally discuss or share things that I am working on or learning about. That will range from lessons learned from working on open source, a cool new tool I discovered, anything I am breaking in my homelab, DevOps workflows, and more.


## About this Blog

This website is built using the following technologies:
- React + Typescript: I had my first experience with React when I started contributing to the [Authelia] open source project. Since then, I have not done any web development in plain HTML, JS, and CSS and prefer to use more structured frameworks.
- Go: Same story, my first exposure to [Go] was working on the Authelia project and my experience thus far has been extremely favorable.
- Tanstack Router & Query: There has been discussion about adopting tanstack components in Authelia due to the developer experience (DEX) benefits. I decided that I might as well learn them since they seem to be tailored to improve DEX. I haven't been disappointed.

This blog is built using the following packages:
- [Grey-matter]: Used for parsing the frontmatter (think metadata) from the blog post files (which are written in markdown). This allows me to define things like post titles, dates, and descriptions in the blog post file and be able to reference it from the frontend.
- [React Markdown]: This is the meat and potatoes, it is a React component that renders Markdown documents into HTML.
- [rehype-highlight]: Provides syntax highlighting for codeblocks using highlight.js. (To help make code blocks pretty)




[Authelia]: https://authelia.com
[Grey-matter]: https://github.com/jonschlinkert/gray-matter
[React Markdown]: https://github.com/remarkjs/react-markdown
[rehype-highlight]: https://github.com/rehypejs/rehype-highlight
[Go]: https://go.dev/