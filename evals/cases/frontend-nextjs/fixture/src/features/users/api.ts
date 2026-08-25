import {UserInput} from "./schema";

export async function createUser(input: UserInput): Promise<{id: string}> {
  const response = await fetch("/api/users", {
    method: "POST",
    headers: {"content-type": "application/json"},
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(`create user failed: ${response.status}`);
  }
  return response.json();
}
