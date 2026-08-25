"use client";

import {createContext, ReactNode, useContext} from "react";

const PermissionContext = createContext<string[]>([]);

export function PermissionProvider({permissions, children}: {permissions: string[]; children: ReactNode}) {
  return <PermissionContext.Provider value={permissions}>{children}</PermissionContext.Provider>;
}

export function useCheckPermission(permission: string): boolean {
  return useContext(PermissionContext).includes(permission);
}
