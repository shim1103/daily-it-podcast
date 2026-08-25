import "@picocss/pico/css/pico.classless.min.css";
import { createPlaybackApiClient } from "./src/api/playback-api-client.ts";
import { createEpisodeListPage } from "./src/pages/episode-list-page.ts";

const baseUrl = "";
const apiClient = createPlaybackApiClient({ baseUrl, fetch: (...args) => fetch(...args) });

function main(): void {
  const app = document.getElementById("app");
  if (!app) {
    return;
  }
  app.replaceChildren(createEpisodeListPage(apiClient, baseUrl));
}

main();
