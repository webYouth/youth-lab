"use client";

// Home page that reads posts from /blog/api/posts.
import { useEffect, useState } from "react";

type Post = {
  id: number;
  title: string;
  content: string;
  created_at: string;
};

export default function HomePage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const run = async () => {
      try {
        const response = await fetch("/blog/api/posts", { cache: "no-store" });
        if (!response.ok) {
          throw new Error(`request failed: ${response.status}`);
        }
        const data = (await response.json()) as { posts?: Post[] };
        setPosts(data.posts || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "unknown error");
      } finally {
        setLoading(false);
      }
    };
    void run();
  }, []);

  return (
    <main>
      <h1>Youth Blog</h1>
      <p>Posts are loaded through Next.js API Route and gRPC.</p>

      {loading && <p>Loading posts...</p>}
      {error && <p>Failed to load posts: {error}</p>}

      {!loading && !error && posts.length === 0 && <p>No posts found.</p>}

      {posts.map((post) => (
        <article key={post.id} className="post">
          <h2>{post.title}</h2>
          <p className="meta">{post.created_at}</p>
          <p>{post.content}</p>
        </article>
      ))}
    </main>
  );
}
