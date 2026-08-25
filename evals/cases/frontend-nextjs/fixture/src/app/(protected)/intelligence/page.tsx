"use client";

import intelligence from "../../../data/intelligence.json";
import {useCheckPermission} from "../../../auth/PermissionContext";
import {PageShell} from "../../../components/PageShell";

export default function IntelligencePage() {
  const canView = useCheckPermission("intelligence:read");
  if (!canView) {
    return <p>Access denied</p>;
  }

  return <PageShell title="Intelligence"><pre>{JSON.stringify(intelligence, null, 2)}</pre></PageShell>;
}
