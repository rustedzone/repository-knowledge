import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export type Session = {
  userID: string;
  permissions: string[];
};

export async function requireSession(): Promise<Session> {
  const token = (await cookies()).get("dashboard_session")?.value;
  if (!token) {
    redirect("/login");
  }

  const [userID, permissions = ""] = token.split(":");
  return {userID, permissions: permissions.split(",").filter(Boolean)};
}
