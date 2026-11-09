import { events, type Event as BroadcastEvent } from "~/mygo";
import { refreshAll } from "~/stores/feeds";
import {
  applyAppearance,
  applySync,
  applySyncSnapshot,
  beginSync,
  handlePostsChanged,
  replaceFeeds,
} from "~/stores/sync-events";
import { openDialog } from "~/stores/ui";

let lastRevision = 0;

/** Follows the Go side's events. Returns the unsubscribe function. */
export function subscribe(): () => void {
  lastRevision = 0;
  return events.appEvent.on((event) => {
    if (event.revision <= lastRevision) return;
    lastRevision = event.revision;
    handleEvent(event);
  });
}

function handleEvent(event: BroadcastEvent): void {
  switch (event.eid) {
    case "state.snapshot":
      if (event.feeds) void replaceFeeds(event.feeds);
      if (event.sync) applySyncSnapshot(event.sync);
      if (event.appearance) applyAppearance(event.appearance.theme, event.appearance.fontSize);
      break;
    case "sync.started":
      if (event.sync) beginSync(event.sync);
      break;
    case "sync.progress":
    case "sync.finished":
      if (event.sync) applySync(event.sync);
      break;
    case "feeds.changed":
      if (event.feeds) void replaceFeeds(event.feeds);
      break;
    case "posts.changed":
      if (event.posts) handlePostsChanged(event.posts.feedId);
      break;
    case "appearance.changed":
      if (event.appearance) applyAppearance(event.appearance.theme, event.appearance.fontSize);
      break;
    case "selection.changed":
      break;
    case "ui.command":
      handleCommand(event.command);
      break;
  }
}

function handleCommand(command: string): void {
  switch (command) {
    case "add-feed":
      openDialog("addFeed");
      break;
    case "refresh":
      void refreshAll();
      break;
    case "settings":
      openDialog("settings");
      break;
  }
}
