import "@picocss/pico/css/pico.classless.min.css";
import { createPlaybackApiClient } from "./src/api/playback-api-client.ts";
import { createEpisodeDetailPage } from "./src/pages/episode-detail-page.ts";
import { createEpisodeListPage } from "./src/pages/episode-list-page.ts";
import { matchRoute } from "./src/utils/match-route.ts";

const apiClient = createPlaybackApiClient({ baseUrl: "", fetch: (...args) => fetch(...args) });

function renderRoute(app: HTMLElement): void {
  const route = matchRoute(window.location.hash);

  if (route.kind === "episode-detail") {
    app.replaceChildren(createEpisodeDetailPage(apiClient, route.episodeId));
    return;
  }

  app.replaceChildren(createEpisodeListPage(apiClient));
}

function main(): void {
  const app = document.getElementById("app");
  if (!app) {
    return;
  }
  window.addEventListener("hashchange", () => renderRoute(app));
  renderRoute(app);
}

main();
