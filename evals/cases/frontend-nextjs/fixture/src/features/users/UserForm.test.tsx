import {userSchema} from "./schema";

it("rejects invalid email addresses", () => {
  expect(userSchema.safeParse({email: "bad", displayName: "Valid"}).success).toBe(false);
});
