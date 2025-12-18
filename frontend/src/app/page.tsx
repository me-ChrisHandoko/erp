import { redirect } from "next/navigation";

export default function Home() {
  // Redirect ke dashboard
  // Nanti bisa diganti dengan redirect ke /login jika belum authenticated
  redirect("/dashboard");
}
