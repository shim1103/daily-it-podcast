import "@picocss/pico/css/pico.classless.min.css";
/* Pico の後: global 夜背景で上書き（Page には書かない） */
import "./src/styles/app.css";
import { createElement } from "react";
import { createRoot } from "react-dom/client";
import { createPlaybackApiClient } from "./src/api/playback-api-client.ts";
import { EpisodeListPage } from "./src/pages/episode-list-page.tsx";

const baseUrl = "";
const apiClient = createPlaybackApiClient({ baseUrl, fetch: (...args) => fetch(...args) });

function main(): void {
  const app = document.getElementById("app");
  if (!app) {
    return;
  }
  createRoot(app).render(createElement(EpisodeListPage, { apiClient, baseUrl }));
}

main();
