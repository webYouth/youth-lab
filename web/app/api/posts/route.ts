// API Route (BFF) that proxies blog posts from gRPC backend.
import { NextResponse } from "next/server";
import { getBlogGrpcClient } from "@/lib/grpc-client";

export const runtime = "nodejs";

type Post = {
  id: number;
  title: string;
  content: string;
  created_at: string;
};

export async function GET() {
  try {
    const client = getBlogGrpcClient();
    const response = await new Promise<{ posts: Post[] }>((resolve, reject) => {
      client.GetPosts({}, (err: Error | null, data: { posts: Post[] }) => {
        if (err) {
          reject(err);
          return;
        }
        resolve(data);
      });
    });

    return NextResponse.json(response, { status: 200 });
  } catch (error) {
    return NextResponse.json(
      {
        posts: [],
        error: error instanceof Error ? error.message : "unknown error"
      },
      { status: 500 }
    );
  }
}
