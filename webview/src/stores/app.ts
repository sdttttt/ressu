/**
 * The public store facade. Components import state and actions from here;
 * the implementations live in the feature modules next to this file.
 */

export { app, currentFeed, feedTitle, isSyncingFeed, visibleFeeds } from "~/stores/app-state";
export type { DialogName } from "~/stores/app-state";
export { messageOf } from "~/stores/errors";
export {
  addFeed,
  exportOPML,
  importOPML,
  refreshAll,
  refreshFeed,
  removeFeed,
  selectFeed,
} from "~/stores/feeds";
export { ensureFeedIcon, forgetFeedIcon, syncFeedIcons } from "~/stores/icons";
export { start } from "~/stores/lifecycle";
export {
  loadMorePosts,
  loadPosts,
  markAllRead,
  PAGE_SIZE,
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
export { openDataDir, openLink, savePaneSizes, updatePrefs } from "~/stores/settings";
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
  showSidebar,
} from "~/stores/ui";
