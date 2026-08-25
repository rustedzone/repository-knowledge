import {z} from "zod";

export const userSchema = z.object({
  email: z.string().email(),
  displayName: z.string().min(2),
});

export type UserInput = z.infer<typeof userSchema>;
