"use client";

import {useFormik} from "formik";
import {useRouter} from "next/navigation";
import {createUser} from "./api";
import {userSchema} from "./schema";

export function UserForm() {
  const router = useRouter();
  const form = useFormik({
    initialValues: {email: "", displayName: ""},
    validate(values) {
      const result = userSchema.safeParse(values);
      if (result.success) return {};
      return result.error.flatten().fieldErrors;
    },
    async onSubmit(values, helpers) {
      try {
        const created = await createUser(values);
        router.push(`/users/${created.id}`);
      } catch (error) {
        helpers.setStatus(error instanceof Error ? error.message : "Unknown error");
      }
    },
  });

  return <form onSubmit={form.handleSubmit}><input name="email" onChange={form.handleChange}/><input name="displayName" onChange={form.handleChange}/><button type="submit">Create</button>{form.status && <p>{form.status}</p>}</form>;
}
