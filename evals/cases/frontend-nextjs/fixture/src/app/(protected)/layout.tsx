import {ReactNode} from "react";
import {PermissionProvider} from "../../auth/PermissionContext";
import {requireSession} from "../../lib/session";

export default async function ProtectedLayout({children}: {children: ReactNode}) {
  const session = await requireSession();
  return <PermissionProvider permissions={session.permissions}>{children}</PermissionProvider>;
}
