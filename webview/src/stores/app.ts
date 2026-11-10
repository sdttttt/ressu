/**
 * The public store facade. Components import state and actions from here;
 * the implementations live in the feature modules next to this file.
 */

export { app, currentFeed, feedTitle, isSyncingFeed, visibleFeeds } from "~/stores/app-state";
export type { DialogName } from "~/stores/app-state";
export { messageOf } from "~/stores/errors";
export {
  exportOPML,
  importOPML,
  openAddFeed,
  refreshAll,
  refreshFeed,
  removeFeed,
  selectFeed,
} from "~/stores/feeds";
export { ensureFeedIcon, syncFeedIcons } from "~/stores/icons";
export { start } from "~/stores/lifecycle";
export {
  loadMorePosts,
  loadPosts,
  markAllRead,
  openImage,
  selectPost,
  setPostSearch,
  setOpenPostRead,
  setPostSortOrder,
  setPostFilter,
  togglePostSearch,
  toggleStar,
} from "~/stores/posts";
export {
  moveSelection,
  openSelectedPost,
  openSelectedPostInBrowser,
  selectNeighbourFeed,
  toggleOpenPostRead,
} from "~/stores/selection";
export { openLink, openSettings, savePaneSizes, updatePrefs } from "~/stores/settings";
export {
  captureSnapshot,
  closeSnapshot,
  copySnapshot,
  saveSnapshot,
} from "~/stores/snapshot-actions";
export { applySync, handlePostsChanged, replaceFeeds } from "~/stores/sync-events";
export {
  closeDialog,
  collapseSidebar,
  expandSidebar,
  hideSidebar,
  keepSidebar,
  openDialog,
  requestSidebar,
  setFeedFilter,
} from "~/stores/ui";
