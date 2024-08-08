import { newAgent } from "./auth";

export async function search(term: string) {
  if (!term.includes(".")) {
    term = `${term}.bsky.social`;
  }
  const agent = newAgent();
  const { data, success } = await agent
    .getProfile({ actor: term.toLowerCase() })
    .catch(() => ({ success: false, data: undefined }));

  if (!success) {
    alert("Could not find user. Please check handle or did, and try again.");
    return;
  }

  useRouter().push(`/users/${data?.did}`);
  return data?.did;
}
