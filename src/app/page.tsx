import { redirect } from "next/navigation";

export default function RootPage() {
  // Temporarily routing to dashboard until we build the (auth)/login page
  redirect("/dashboard");
}