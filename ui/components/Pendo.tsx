import React, { useContext } from "react";
import { shake128 } from "js-sha3";
import Mnemonic from "mnemonic-browser";
import { Auth } from "../contexts/AuthContext";
import { CoreClientContext } from "../contexts/CoreClientContext";

declare global {
  interface Window {
    pendo: any;
  }
}

const key = "VyzGoWoKvtJHyTnU+GVhDe+wU9bwZDH87bp505/0f/2UIpHzB+tmyZmfsH8/iJoH";

export interface Props {
  /** Value to use as default if the telemetry flag cannot be read. */
  defaultTelemetryFlag: string;
}

export default function Pendo({ defaultTelemetryFlag }: Props) {
  const { featureFlags: flags } = useContext(CoreClientContext);
  const { userInfo } = useContext(Auth);
  const [isPendoInitialized, setIsPendoInitialized] = React.useState(false);
  const [isPendoAgentReady, setIsPendoAgentReady] = React.useState(false);

  const telemetryFlag =
    flags?.WEAVE_GITOPS_FEATURE_TELEMETRY || defaultTelemetryFlag;

  const shouldInitPendo = !!flags && telemetryFlag === "true";

  React.useEffect(() => {
    if (!shouldInitPendo) {
      return;
    }

    let visitorId = "";
    const userEmail = userInfo?.email;

    if (userEmail) {
      const hasher = shake128.create(128);
      hasher.update(key);
      hasher.update(userEmail);
      visitorId = Mnemonic.fromHex(hasher.hex()).toWords().join("-");
    }

    const visitor = {
      id: visitorId,
    };

    const accountId = Mnemonic.fromHex(flags.ACCOUNT_ID).toWords().join("-");

    // This is copied from the pendo docs
    // eslint unwrapps it, and the initialize call has been customized
    /* eslint-disable */
    (function (apiKey) {
      if (isPendoInitialized) {
        let shouldIdentify = true;

        if (isPendoAgentReady) {
          const currentVisitorId = window.pendo.getVisitorId();
          const currentAccountId = window.pendo.getAccountId();

          if (
            currentVisitorId === visitorId &&
            currentAccountId === accountId
          ) {
            shouldIdentify = false;
          }
        }

        if (shouldIdentify) {
          window.pendo.identify(visitorId, accountId);
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

          account: {
            id: accountId,
            devMode: flags.WEAVE_GITOPS_FEATURE_DEV_MODE === "true",
          },

          events: {
            ready: function () {
              setIsPendoAgentReady(true);
            },
          },
        });

        setIsPendoInitialized(true);
      }
    })("7a83d612-fa5b-4bfe-4544-861a89ceaf89");
  }, [shouldInitPendo, userInfo]);

  return <></>;
}
