import { toast } from "solid-sonner";

import { Posts } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { closeSnapshot } from "~/stores/snapshot-actions";

/** How many posts a page of the list holds. */
export const PAGE_SIZE = 100;

/** Loads a page of the selected feed's posts, replacing or appending them. */
export async function loadPosts(offset = 0): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) {
    setState({ posts: [], postTotal: 0, postUnread: 0 });
    return;
  }

  setState("loadingPosts", true);
  try {
    const page = await Posts.list({
      feedId,
      unreadOnly: state.unreadOnly,
      starredOnly: state.starredOnly,
      search: "",
      offset,
      limit: PAGE_SIZE,
    });
    if (offset === 0) setState("posts", page.posts);
    else setState("posts", (posts) => [...posts, ...page.posts]);
    setState({ postTotal: page.total, postUnread: page.unread });
  } catch (error) {
    toast.error(`无法读取文章：${messageOf(error)}`);
  } finally {
    setState("loadingPosts", false);
  }
}

/** Loads the next page of posts. */
export async function loadMorePosts(): Promise<void> {
  if (state.loadingPosts) return;
  await loadPosts(state.posts.length);
}

/** Filters the list by unread or starred posts. */
export async function setPostFilter(filter: {
  unreadOnly?: boolean;
  starredOnly?: boolean;
}): Promise<void> {
  setState({
    unreadOnly: filter.unreadOnly ?? state.unreadOnly,
    starredOnly: filter.starredOnly ?? state.starredOnly,
  });
  await loadPosts(0);
}

/** Opens a post, and marks it read when the preferences ask for that. */
export async function selectPost(guid: string): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) return;

  closeSnapshot();
  setState("current", "guid", guid);
  setState("loadingPost", true);
  try {
    const post = await Posts.get(feedId, guid);
    setState("post", post);
    if (state.prefs.markReadOnOpen && !post.read) {
      await Posts.markRead(feedId, guid, true);
      setReadLocally(guid, true);
    }
  } catch (error) {
    toast.error(`无法打开文章：${messageOf(error)}`);
  } finally {
    setState("loadingPost", false);
  }
}

/** Marks the open post read or unread. */
export async function setOpenPostRead(read: boolean): Promise<void> {
  const feedId = state.current.feedId;
  const post = state.post;
  if (feedId === null || !post) return;

  try {
    await Posts.markRead(feedId, post.guid, read);
    setReadLocally(post.guid, read);
    setState("post", "read", read);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Marks every post of a feed read, the selected one by default. */
export async function markAllRead(feedId: number | null = state.current.feedId): Promise<void> {
  if (feedId === null) return;

  try {
    const changed = await Posts.markAllRead(feedId);
    if (changed === 0) {
      toast.info("这个订阅源没有未读文章");
      return;
    }
    toast.success(`已把 ${changed} 篇文章标记为已读`);
    if (feedId === state.current.feedId) await loadPosts(0);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Stars or unstars a post. */
export async function toggleStar(guid: string): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) return;

  try {
    const updated = await Posts.toggleStar(feedId, guid);
    setState("posts", (post) => post.guid === guid, "starred", updated.starred);
    if (state.post?.guid === guid) setState("post", "starred", updated.starred);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

function setReadLocally(guid: string, read: boolean): void {
  setState("posts", (post) => post.guid === guid, "read", read);
  setState("postUnread", (unread) => Math.max(read ? unread - 1 : unread + 1, 0));
}
