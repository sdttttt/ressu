import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("~/mygo", () => ({
  Status: {
    notify: vi.fn(),
  },
}));

import { toast } from "~/lib/notifications";
import { Status } from "~/mygo";

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(Status.notify).mockResolvedValue(undefined);
});

describe("toast", () => {
  it("forwards each level through Status.notify", () => {
    toast.error("error");
    toast.warning("warning");
    toast.info("info");
    toast.success("success");

    expect(Status.notify).toHaveBeenNthCalledWith(1, "error", "error");
    expect(Status.notify).toHaveBeenNthCalledWith(2, "warning", "warning");
    expect(Status.notify).toHaveBeenNthCalledWith(3, "info", "info");
    expect(Status.notify).toHaveBeenNthCalledWith(4, "success", "success");
  });

  it("does not surface native notification failures", async () => {
    vi.mocked(Status.notify).mockRejectedValue(new Error("window closed"));

    expect(() => toast.error("message")).not.toThrow();
    await Promise.resolve();
  });
});
