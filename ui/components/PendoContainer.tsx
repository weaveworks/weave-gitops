import React from "react";
import p from "../../package.json";

import { useVersion } from "../hooks/version";
import { GetVersionResponse } from "../lib/api/core/core.pb";
import { getAppVersion } from "../lib/utils";
import Pendo from "./Pendo";
import { noVersion } from "./Version";

export const tier = "oss";

/** This component is used to wrap the Pendo component and pass in the version.
 * Use it where you are authorized to make API calls
 * and need to track the app version with Pendo.
 */
export default function PendoContainer() {
  const [version, setVersion] = React.useState<string>("");
  const [shouldWaitForVersion, setShouldWaitForVersion] =
    React.useState<boolean>(true);

  const { data, isLoading } = useVersion();
  const versionData = data || ({} as GetVersionResponse);

  React.useEffect(() => {
    if (isLoading) return;

    if (!versionData) {
      setShouldWaitForVersion(false);
      return;
    }

    const appVersion = getAppVersion(data, p.version, isLoading);

    setVersion(appVersion.versionText || noVersion);
  }, [versionData, isLoading]);

  return (
    <Pendo
      defaultTelemetryFlag="false"
      tier={tier}
      shouldWaitForVersion={shouldWaitForVersion}
      version={version}
    />
  );
}
