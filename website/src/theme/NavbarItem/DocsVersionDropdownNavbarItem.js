import React from "react";
import DocsVersionDropdownNavbarItem from '@theme-original/NavbarItem/DocsVersionDropdownNavbarItem';
import { useLocation }  from '@docusaurus/router';

export default function DocsVersionDropdownNavbarItemWrapper(props) {
  const { docsPluginId, className, type } = props
  const { pathname } = useLocation()

  /*
  custom navbar fix to show the correct version dropdown for either oss or ee docs.
  docs which are split like this are given 'docsPluginId's.
  so here we check that the plugin id matches some part of the url path and display
  the right version dropdown in the navbar.

  docusaurus does advise against wrapping the navbar component, so....

  the docsPluginId is set in the docusaurus.config.js under navbar

  {
    type: "docsVersionDropdown",
    position: "right",
    dropdownActiveClassDisabled: true,
    docsPluginId: "default",
  },
  {
    type: "docsVersionDropdown",
    position: "right",
    dropdownActiveClassDisabled: true,
    docsPluginId: "enterprise",
  },

  annoyingly you are not allowed to rename the preexisting one to 'docs', because
  there always has to be a 'default'. this means that
  thing i am doing below to check whether the docsPluginId is in the url path wont
  work unless i reset the routeBasePath, which i don't want to because i would then
  need to redirect all our urls. so i am leaving everything as it is and doing an
  extra check in the code here.
  */

  const doesPathnameContainDocsPluginId = pathname.includes(docsPluginId)
  const isDefaultDocs = (pathname.includes('docs') && docsPluginId == "default")

  if (!doesPathnameContainDocsPluginId && !isDefaultDocs) {
    return null
  }

  return <DocsVersionDropdownNavbarItem {...props} />;
}
