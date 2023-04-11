import _ from "lodash";
import { DateTime } from "luxon";
import { toast } from "react-toastify";
import styled from "styled-components";
import Flex from "../components/Flex";
import { computeReady, ReadyType } from "../components/KubeStatusIndicator";
import { AppVersion, repoUrl } from "../components/Version";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { Condition, Kind, ObjectRef } from "./api/core/types.pb";
import { Automation, HelmRelease, Kustomization } from "./objects";

export function notifySuccess(message: string) {
  toast["success"](message);
}

export function notifyError(message: string) {
  toast["error"](`Error: ${message}`);
}

export function poller(cb, interval): any {
  if (process.env.NODE_ENV === "test") {
    // Stay synchronous in tests
    return cb();
  }

  return setInterval(cb, interval);
}

// isHTTP checks if something looks link-like enough that we think it
// would be a good idea to auto-link. This is quite strict, and does
// not allow e.g relative links. See also isAllowedLink
export function isHTTP(uri: string): boolean {
  // Regex from Diego Perini's gist: https://gist.github.com/dperini/729294
  // It works better than other regular expressions for validating HTTP and HTTPS URLs.
  const regex = new RegExp(
    "^" +
      // protocol identifier (optional)
      // short syntax // still required
      "(?:(?:(?:https?):)?\\/\\/)" +
      // user:pass BasicAuth (optional)
      "(?:\\S+(?::\\S*)?@)?" +
      "(?:" +
      // IP address dotted notation octets
      // excludes loopback network 0.0.0.0
      // excludes reserved space >= 224.0.0.0
      // excludes network & broadcast addresses
      // (first & last IP address of each class)
      "(?:[1-9]\\d?|1\\d\\d|2[01]\\d|22[0-3])" +
      "(?:\\.(?:1?\\d{1,2}|2[0-4]\\d|25[0-5])){2}" +
      "(?:\\.(?:[1-9]\\d?|1\\d\\d|2[0-4]\\d|25[0-4]))" +
      "|" +
      // host & domain names, may end with dot
      // can be replaced by a shortest alternative
      // (?![-_])(?:[-\\w\\u00a1-\\uffff]{0,63}[^-_]\\.)+
      "(?:" +
      "(?:" +
      "[a-z0-9\\u00a1-\\uffff]" +
      "[a-z0-9\\u00a1-\\uffff_-]{0,62}" +
      ")?" +
      "[a-z0-9\\u00a1-\\uffff]\\." +
      ")+" +
      // TLD identifier name, may end with dot
      "(?:[a-z\\u00a1-\\uffff]{2,}\\.?)" +
      ")" +
      // locahost is the only name allowed without TLD
      "|" +
      "localhost" +
      // port number (optional)
      "(?::\\d{2,5})?" +
      // resource path (optional)
      "(?:[/?#]\\S*)?" +
      "$",
    "i"
  );

  return regex.test(uri);
}

// isAllowedLink checks if making a "link" clickable will do anybody
// any good. This is quite permissive - it's to stop e.g. oci:// links
// being clickable, because clicking them will make nobody happy. See
// also isAllowedLink
export function isAllowedLink(uri: string): boolean {
  // Regex from https://github.com/cure53/DOMPurify/blob/cce00ac40d33c2aae6422eaa59e6a8aad5c73901/src/regexp.js
  const regex = new RegExp(
    /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|cid|xmpp):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i // eslint-disable-line no-useless-escape
  );
  return regex.test(uri);
}

export function convertGitURLToGitProvider(uri: string): string {
  if (isHTTP(uri)) {
    return uri;
  }

  const matches = uri.match(/git@(.*)[/|:](.*)\/(.*)/);
  if (!matches) {
    return "";
  }
  const [, provider, org, repo] = matches;

  return `https://${provider}/${org}/${repo}`;
}

export function pageTitleWithAppName(title: string, appName?: string): string {
  return `${title}${appName ? ` for ${appName}` : ""}`;
}

interface Statusable {
  conditions: Condition[];
  suspended: boolean;
}

export function statusSortHelper({
  suspended,
  conditions,
}: Statusable): number {
  if (suspended) return 2;
  if (computeReady(conditions) === ReadyType.Reconciling) return 3;
  else if (computeReady(conditions) === ReadyType.Ready) return 4;
  else return 1;
}

export function automationLastUpdated(a: Kustomization | HelmRelease): string {
  return _.get(_.find(a?.conditions, { type: "Ready" }), "timestamp");
}

export function makeImageString(images: string[]): string {
  let imageString = "";
  if (!images[0]) return "-";
  else imageString += images[0];
  if (images[1]) {
    for (let i = 1; i < images.length; i++) imageString += `\n${images[i]}`;
  }
  return imageString;
}

export function formatMetadataKey(key: string): string {
  return key
    .replace(/-/g, " ")
    .replace(/\w+/g, (w) => w[0].toUpperCase() + w.slice(1));
}

export const convertImage = (image: string) => {
  const split = image.split("/");

  //remove tags
  const tag = split[split.length - 1];
  if (tag.includes(":"))
    split[split.length - 1] = tag.slice(0, tag.indexOf(":"));

  const prefix = split.shift();
  const noTag = split.join("/");
  let url = "";

  //Github GHCR or Google GCR
  if (prefix === "ghcr.io" || prefix === "gcr.io")
    return `https://${prefix}/${noTag}`;
  //Quay.io
  if (prefix === "quay.io") {
    return `https://quay.io/repository/${noTag}`;
  }
  //complex docker prefix case
  if (prefix === "docker.io") {
    url = "https://hub.docker.com/r/";
    //library alias
    if (split[0] === "library") return url + "_/" + split[1];
    //global
    if (!split[1]) return url + "_/" + split[0];
    //namespaced
    return url + noTag;
  }
  //docker without prefix
  if (prefix === "library") return "https://hub.docker.com/r/_/" + split[0];
  //this one's at risk if we have to add others - global docker images can just be one word apparently
  if (!split[0]) {
    return "https://hub.docker.com/r/_/" + prefix;
  }
  //any other url
  if (prefix.includes(".")) return false;
  //one slash docker images w/o docker.io
  return `https://hub.docker.com/r/${prefix}/${noTag}`;
};

// getSourceRefForAutomation returns the automation's sourceRef
// depending on whether the automation is a Kustomization or a HelmRelease.
export function getSourceRefForAutomation(
  automation?: Automation
): ObjectRef | undefined {
  return automation?.type === Kind.Kustomization
    ? (automation as Kustomization)?.sourceRef
    : (automation as HelmRelease)?.helmChart?.sourceRef;
}

// getAppVersion returns the app version to display in the UI or track in analytics.
export function getAppVersion(
  versionData: GetVersionResponse,
  defaultVersion: string,
  isLoading = false,
  defaultVersionPrefix = ""
): AppVersion {
  const shouldDisplayApiVersion =
    !isLoading &&
    (versionData?.semver || "").replace(/^v+/, "") !== defaultVersion &&
    versionData?.branch &&
    versionData?.commit;

  const versionText = shouldDisplayApiVersion
    ? `${versionData.branch}-${versionData.commit}`
    : `${defaultVersionPrefix}${defaultVersion}`;
  const versionHref = shouldDisplayApiVersion
    ? `${repoUrl}/commit/${versionData.commit}`
    : `${repoUrl}/releases/tag/v${defaultVersion}`;

  return {
    versionText,
    versionHref,
  };
}

// formatLogTimestamp formats a timestamp string in the RFC3339 format
// to a human-readable format with UTC offset.
// If the timestamp is undefined or an empty string, it returns "-".
export function formatLogTimestamp(timestamp?: string, zone?: string): string {
  if (!timestamp) {
    return "-";
  }

  let dt = DateTime.fromISO(timestamp);

  if (zone) {
    dt = dt.setZone(zone);
  }

  let formattedTimestamp = `${dt.toFormat("yyyy-LL-dd HH:mm:ss 'UTC'Z")}`;

  if (dt.offset === 0) {
    formattedTimestamp = formattedTimestamp.replace("UTC+0", "UTC");
  }

  return formattedTimestamp;
}

export const createYamlCommand = (
  kind: string,
  name: string,
  namespace: string
): string => {
  if (kind && name) {
    const namespaceString = namespace ? ` -n ${namespace}` : "";
    return `kubectl get ${kind.toLowerCase()} ${name}${namespaceString} -o yaml`;
  }
  return null;
};

export const Fade = styled(Flex)<{
  fade: boolean;
}>`
  opacity: ${({ fade }) => (fade ? 0 : 1)};
  transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
  ${({ fade }) => fade && "pointer-events: none"};
`;
