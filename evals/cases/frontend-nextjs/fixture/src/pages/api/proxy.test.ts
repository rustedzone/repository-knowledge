import handler from "./[...paths]";

it("exports the catch-all BFF handler", () => {
  expect(handler).toBeDefined();
});
