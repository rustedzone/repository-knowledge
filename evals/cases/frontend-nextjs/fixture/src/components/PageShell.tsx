import {ReactNode} from "react";

export function PageShell({title, children}: {title: string; children: ReactNode}) {
  return <main className="mx-auto max-w-6xl p-6"><h1 className="text-2xl font-semibold">{title}</h1>{children}</main>;
}
