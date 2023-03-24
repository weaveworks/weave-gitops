import React, { useContext } from "react";
import { shake128 } from "js-sha3";
import Mnemonic from "mnemonic-browser";
import { Auth } from "../contexts/AuthContext";
import { CoreClientContext } from "../contexts/CoreClientContext";
import { noVersion } from "../components/Version";
import { useFeatureFlags } from "../hooks/featureflags";

declare global {
  interface Window {
    pendo: any;
  }
}

const pendoKey = "pendo";
const hashKey =
  "VyzGoWoKvtJHyTnU+GVhDe+wU9bwZDH87bp505/0f/2UIpHzB+tmyZmfsH8/iJoH";

export interface Props {
  /** Value to use as default if the telemetry flag cannot be read. */
  defaultTelemetryFlag: string;
  /** Dashboard tier */
  tier: string;
  /** If this flag is set to true, wait for the version prop value before initializing Pendo. */
  shouldWaitForVersion?: boolean;
  /** Dashboard version */
  version?: string;
}

export default function Pendo({
  defaultTelemetryFlag,
  tier,
  shouldWaitForVersion,
  version,
}: Props) {
  const { featureFlags: flags } = useContext(CoreClientContext);
  const { isFlagEnabled } = useFeatureFlags();

  const { userInfo } = useContext(Auth);
  const [isPendoInitialized, setIsPendoInitialized] = React.useState(false);
  const [isPendoAgentReady, setIsPendoAgentReady] = React.useState(false);

  React.useEffect(() => {
    const telemetryFlag =
      isFlagEnabled("WEAVE_GITOPS_FEATURE_TELEMETRY") || defaultTelemetryFlag;

    const shouldInitPendo =
      !!flags &&
      telemetryFlag === "true" &&
      (!shouldWaitForVersion || !!version);

    if (!shouldInitPendo) {
      return;
    }

    let pendoKeys: string[] = [];

    if (!window.localStorage) {
      console.warn("no local storage found");
    } else {
      pendoKeys = Object.keys(window.localStorage).filter(
        (key) => key.toLowerCase().indexOf(pendoKey) != -1
      );
    }

    const userEmail = userInfo?.email;

    if (!userEmail && pendoKeys.length == 0) {
      return;
    }

    let visitorId = "";

    if (userEmail) {
      const hasher = shake128.create(128);
      hasher.update(hashKey);
      hasher.update(userEmail);
      visitorId = Mnemonic.fromHex(hasher.hex()).toWords().join("-");
    }

    const shouldAddVersion = !!version && version !== noVersion;

    const visitor = {
      id: visitorId,
      tier,
      ...(shouldAddVersion && { latestVersion: version }),
    };

    const accountId = Mnemonic.fromHex(isFlagEnabled("ACCOUNT_ID"))
      .toWords()
      .join("-");

    const account = {
      id: accountId,
      devMode: isFlagEnabled("WEAVE_GITOPS_FEATURE_DEV_MODE"),
    };

    // This is copied from the pendo docs
    // eslint unwrapps it, and the initialize call has been customized
    /* eslint-disable */
    (function (apiKey) {
      if (isPendoInitialized) {
        let shouldIdentify = true;

        if (isPendoAgentReady) {
          const currentVisitorId = window.pendo.getVisitorId();
          const currentAccountId = window.pendo.getAccountId();
          const metadata = window.pendo.getSerializedMetadata();

          if (
            currentVisitorId === visitorId &&
            currentAccountId === accountId &&
            metadata?.latestVersion === version
          ) {
            shouldIdentify = false;
          }
        }

        if (shouldIdentify) {
          window.pendo.identify({ visitor, account });
        }
      } else {
        (function (p, e, n, d, o) {
          let v, w, x, y, z;
          o = p[d] = p[d] || {};
          o._q = o._q || [];
          v = ["initialize", "identify", "updateOptions", "pageLoad", "track"];
          for (w = 0, x = v.length; w < x; ++w)
            (function (m) {
              o[m] =
                o[m] ||
                function () {
                  o._q[m === v[0] ? "unshift" : "push"](
                    [m].concat([].slice.call(arguments, 0))
                  );
                };
            })(v[w]);
          y = e.createElement(n);
          y.async = !0;
          y.src = "https://cdn.pendo.io/agent/static/" + apiKey + "/pendo.js";
          z = e.getElementsByTagName(n)[0];
          z.parentNode.insertBefore(y, z);
        })(window, document, "script", "pendo");

        window.pendo.initialize({
          visitor: visitor,

          account: account,

          events: {
            ready: function () {
              setIsPendoAgentReady(true);
            },
          },
        });

        setIsPendoInitialized(true);
      }
    })("7a83d612-fa5b-4bfe-4544-861a89ceaf89");
  }, [flags, userInfo, shouldWaitForVersion, version]);

  return <></>;
}
