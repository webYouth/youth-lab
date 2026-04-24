// Root layout for the Next.js App Router.
import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Youth Blog",
  description: "Personal blog frontend powered by Next.js and gRPC BFF"
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
