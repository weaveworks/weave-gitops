import React, { useContext } from "react";
import { shake128 } from "js-sha3";
import Mnemonic from "mnemonic-browser";
import { useFeatureFlags } from "../hooks/featureflags";
import { Auth } from "../contexts/AuthContext";

declare global {
  interface Window {
    pendo: any;
  }
}

export default function Pendo() {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};
  const { userInfo } = useContext(Auth);

  if (flags.WEAVE_GITOPS_FEATURE_TELEMETRY == "true") {
    if (!userInfo || !userInfo.email) {
      return <></>;
    }

    const key =
      "VyzGoWoKvtJHyTnU+GVhDe+wU9bwZDH87bp505/0f/2UIpHzB+tmyZmfsH8/iJoH";

    const hasher = shake128.create(128);
    hasher.update(key);
    hasher.update(userInfo.email);

    const visitorId = Mnemonic.fromHex(hasher.hex()).toWords().join("-");
    const accountId = Mnemonic.fromHex(flags.ACCOUNT_ID).toWords().join("-");

    // This is copied from the pendo docs
    // eslint unwrapps it, and the initialize call has been customized
    /* eslint-disable */
    (function (apiKey) {
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
        visitor: {
          id: visitorId,
        },

        account: {
          id: accountId,
          devMode: flags.WEAVE_GITOPS_FEATURE_DEV_MODE == "true",
        },
      });
    })("7a83d612-fa5b-4bfe-4544-861a89ceaf89");
  }

  return <></>;
}
