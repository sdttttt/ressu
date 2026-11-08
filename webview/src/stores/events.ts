import { events } from "~/mygo";
import { refreshAll } from "~/stores/feeds";
import { applySync, handlePostsChanged, replaceFeeds } from "~/stores/sync-events";
import { openDialog } from "~/stores/ui";

/** Follows the Go side's events. Returns the unsubscribe function. */
export function subscribe(): () => void {
  const offs = [
    events.feedsChanged.on((snapshot) => void replaceFeeds(snapshot.feeds)),
    events.postsChanged.on((change) => handlePostsChanged(change.feedId)),
    events.syncState.on((state) => applySync(state)),
    events.uiCommand.on((command) => handleCommand(command)),
  ];

  return () => {
    for (const off of offs) off();
  };
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
