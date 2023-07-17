import React from 'react';
import {
  useVersions,
  useActiveDocContext,
} from '@docusaurus/plugin-content-docs/client';
import {useLocation} from '@docusaurus/router';

const getVersionMainDoc = (version) =>
  version.docs.find((doc) => doc.id === version.mainDocId);
  
const VersionItems = () => {
  const {search, hash} = useLocation();
  const activeDocContext = useActiveDocContext('default');
  const versions = useVersions('default');
  
  const versionLinks = versions.filter(function(version) {
    if (version.isLast || version.label === "Next") { // omit current and lastVersion from the list
      return false; // skip
    }
    return true;
  }).map((version) => {
    // We try to link to the same doc, in another version
    // When not possible, fallback to the "main doc" of the version
    const versionDoc =
      activeDocContext.alternateDocVersions[version.name] ??
      getVersionMainDoc(version);
    return {
      label: version.label,
      // preserve ?search#hash suffix on version switches
      to: `${versionDoc.path}${search}${hash}`,
    };
  });

  return (
    <ul>
    {versionLinks.map(data => (
            <li><a href={data.to}>{data.label}</a></li>
        ))}
    </ul>   
  );
}

export default VersionItems