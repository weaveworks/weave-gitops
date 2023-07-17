import React from 'react';
import Layout from '@theme/Layout';
import VersionItems from '@site/src/components/NavbarItem/VersionItems'
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

export default function Home() {
  const {siteConfig} = useDocusaurusContext();

  return (
    <Layout 
      title={`${siteConfig.title}`}
      description={`${siteConfig.tagline}`}>
      <main class="container container--fluid margin-vert--lg">
        <div class="row">
            <div class="col">
                <article>
                    <h1>Version Archives</h1>
                    <p>Snapshots of the <a href="/">documentation</a> website for each past release.</p>
                    <div class="dropdown_links"><VersionItems /></div>
                </article>
            </div>
        </div>
        
      </main>
    </Layout>
  );
}
