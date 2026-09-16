/** remote TEST R2 binding 起動用の最小 Worker。binding 取得だけが目的。 */
export default {
  async fetch() {
    return new Response("playback-smoke-r2-peer");
  },
};
