import { cleanup, fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { afterEach, describe, expect, it, vi } from "vitest";

import AddFeedDialog from "~/components/add-feed-dialog";

const preview = {
  title: "C 站",
  siteUrl: "https://c.example/",
  description: "",
  iconUrl: "",
  language: "",
};

const fixtures = vi.hoisted(() => ({
  addFeed: vi.fn(),
  testFeed: vi.fn(),
}));

vi.mock("~/stores/app", () => ({
  app: { dialog: { addFeed: true } },
  addFeed: fixtures.addFeed,
  closeDialog: vi.fn(),
  importOPML: vi.fn(),
  openDialog: vi.fn(),
  testFeed: fixtures.testFeed,
}));

describe("AddFeedDialog", () => {
  afterEach(() => cleanup());

  it("fills an empty name after a successful test and enables adding", async () => {
    fixtures.testFeed.mockResolvedValue({ feed: preview, error: "" });
    fixtures.addFeed.mockResolvedValue(true);
    render(() => <AddFeedDialog />);

    const addButton = screen.getByRole("button", { name: "添加" }) as HTMLButtonElement;
    expect(addButton.disabled).toBe(true);

    fireEvent.input(screen.getByLabelText("订阅地址"), {
      target: { value: "https://c.example/feed" },
    });
    fireEvent.click(screen.getByRole("button", { name: "测试" }));

    await waitFor(() => expect(screen.getByLabelText("名称")).toHaveProperty("value", "C 站"));
    expect(screen.getByText("测试通过，可以添加。")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "添加" }));
    expect(fixtures.addFeed).toHaveBeenCalledWith("C 站", "https://c.example/feed");
  });

  it("keeps adding disabled when testing fails", async () => {
    fixtures.testFeed.mockResolvedValue({ feed: null, error: "无法读取订阅源" });
    render(() => <AddFeedDialog />);

    fireEvent.input(screen.getByLabelText("订阅地址"), {
      target: { value: "https://c.example/feed" },
    });
    fireEvent.click(screen.getByRole("button", { name: "测试" }));

    await waitFor(() => expect(screen.getByText("无法读取订阅源")).toBeTruthy());
    expect((screen.getByRole("button", { name: "添加" }) as HTMLButtonElement).disabled).toBe(true);
    expect(fixtures.addFeed).not.toHaveBeenCalled();
  });

  it("requires testing again after the address changes", async () => {
    fixtures.testFeed.mockResolvedValue({ feed: preview, error: "" });
    render(() => <AddFeedDialog />);
    const url = screen.getByLabelText("订阅地址");

    fireEvent.input(url, { target: { value: "https://c.example/feed" } });
    fireEvent.click(screen.getByRole("button", { name: "测试" }));
    await waitFor(() => expect(fixtures.testFeed).toHaveBeenCalledTimes(1));

    fireEvent.input(url, { target: { value: "https://d.example/feed" } });

    expect((screen.getByRole("button", { name: "添加" }) as HTMLButtonElement).disabled).toBe(true);
  });
});
