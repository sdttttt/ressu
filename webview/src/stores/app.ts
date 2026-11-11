/**
 * The public store facade. Components import state and actions from here;
 * the implementations live in the feature modules next to this file.
 */

export { app, currentFeed, feedTitle } from "~/stores/app-state";
export type { DialogName } from "~/stores/app-state";
export { openAddFeed, refreshAll } from "~/stores/feeds";
export { start } from "~/stores/lifecycle";
export {
  loadMorePosts,
  markAllRead,
  openImage,
  selectNeighbourPost,
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
export { openLink, openSettings } from "~/stores/settings";
export {
  captureSnapshot,
  closeSnapshot,
  copySnapshot,
  saveSnapshot,
} from "~/stores/snapshot-actions";
export { closeDialog, openDialog } from "~/stores/ui";
