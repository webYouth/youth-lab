// Singleton gRPC client for BlogService.
import path from "node:path";
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";

const protoPath = path.join(process.cwd(), "proto", "blog.proto");
const packageDefinition = protoLoader.loadSync(protoPath, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true
});

const protoDescriptor = grpc.loadPackageDefinition(packageDefinition) as any;
const BlogService = protoDescriptor.blog.v1.BlogService;

let clientInstance: grpc.Client | null = null;

export function getBlogGrpcClient() {
  if (clientInstance) {
    return clientInstance as any;
  }

  const target = process.env.GRPC_SERVER_ADDR || "localhost:50051";
  clientInstance = new BlogService(target, grpc.credentials.createInsecure());
  return clientInstance as any;
}
