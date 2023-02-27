import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faListCheck, faKey, faCubes, faPersonRunning } from '@fortawesome/free-solid-svg-icons'

const CoreFeatureList = [
  {
    title: 'Application Management',
    icon: faListCheck,
    description: (
      <>
        Understand and manage application lifecycle in a GitOps enabled cluster.
        Immediately detect drift and evaluate cluster health, even perform roll back actions
        as well as monitor continuous operations.
      </>
    ),
  },
  {
    title: 'Enterprise Integration',
    icon: faKey,
    description: (
      <>
        Weave GitOps easily integrates into enterprise SSO to provide fine grained
        roles based access control (RBAC) to clusters and their assets.
      </>
    ),
  },
  {
    title: 'Terraform + GitOps',
    icon: faCubes,
    description: (
      <>
        Bring the power of GitOps to Terraform management: reconciliation, automation,
        drift detection, and more. GitOps-ify infrastructure and application resources
        at your own pace, adding capabilities as you progress.
      </>
    ),
  },
  {
    title: 'GitOps Run',
    icon: faPersonRunning,
    description: (
      <>
        An easy way to get started with GitOps that shows reconciliations and
        issues in real time, shortening troubleshooting time.
      </>
    ),
  },
];

function Feature({icon, title, description}) {
  return (
    <div className={clsx('col col--6')}>
      <div className="text--center">
        <FontAwesomeIcon className={styles.featureSvg} role="img" icon={icon} />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function CoreFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {CoreFeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
