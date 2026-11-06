import { events } from "~/mygo";
import { applySync, handlePostsChanged, openDialog, refreshAll, replaceFeeds } from "~/stores/app";

/** Follows the Go side's events. Returns the unsubscribe function. */
export function subscribe(): () => void {
  const offs = [
    events.feedsChanged.on((snapshot) => replaceFeeds(snapshot.feeds)),
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
