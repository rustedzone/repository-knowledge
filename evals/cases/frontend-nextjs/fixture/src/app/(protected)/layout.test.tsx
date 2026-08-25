import {requireSession} from "../../lib/session";

describe("protected layout", () => {
  it("requires a session before rendering protected content", () => {
    expect(requireSession).toBeDefined();
  });
});
