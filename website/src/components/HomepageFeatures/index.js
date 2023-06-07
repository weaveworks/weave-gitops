import React from 'react';
import Link from '@docusaurus/Link';
import CoreFeatures from './../CoreFeatures';
import EEFeatures from './../EnterpriseFeatures';
import styles from './styles.module.css';
import ReactPlayer from "react-player/lazy";
import clsx from 'clsx';

export default function HomepageFeatures() {
  return (
    <div>
      <div className="container text--center margin-bottom--xl margin-top--lg">
        <h2>
        Weave GitOps is a powerful extension to <Link to="https://fluxcd.io">Flux</Link>,
        a leading GitOps engine and CNCF project.
        </h2>

        <h2>
        Weave GitOps provides insights into your
        application deployments, and makes continuous delivery with GitOps easier to adopt
        and scale across your teams.
        </h2>
        <h1>See GitOps in action!</h1>
        <div className={styles.playerWrapper}>
          <ReactPlayer
            class={styles.reactPlayer}
            controls="true" width="100%" height="100%"
            url="https://www.youtube.com/watch?v=2TJz7RhDtAc" />
        </div>
      </div>
      <div className="text--center padding-horiz--md">
        <h1>Main Features</h1>
        <h3>
            Available on <Link to="/docs/intro-weave-gitops">all Tiers</Link>
        </h3>
      </div>
      <CoreFeatures/>

      <div className="text--center padding-horiz--md">
        <h1>Enterprise Features</h1>
        <h3>
            Available only on the <Link to="/docs/enterprise/getting-started">Enterprise Tier</Link>
        </h3>
      </div>
      <EEFeatures col={clsx('col--4')}/>
    </div>
  );
}
