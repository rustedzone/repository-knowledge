"use client";

import useSWR from "swr";

const fetcher = async (url: string) => {
  const response = await fetch(url);
  if (!response.ok) throw new Error(`users request failed: ${response.status}`);
  return response.json();
};

export function useUsers() {
  return useSWR("/api/users", fetcher);
}
