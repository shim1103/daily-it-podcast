/** local S3 peer 用の最小 Worker。binding 起動だけが目的。 */
export default {
  async fetch() {
    return new Response("r2locals3-peer");
  },
};
