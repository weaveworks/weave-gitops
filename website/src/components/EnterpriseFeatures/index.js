import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faSailboat,
  faHandshake,
  faBarsProgress,
  faHandPointer,
  faHeartPulse,
  faUsers,
  faMapLocationDot,
  faBell,
  faInfinity,
} from '@fortawesome/free-solid-svg-icons'

const EEFeatureList = [
  {
    title: 'Cluster Fleet Management',
    icon: faSailboat,
    description: (
      <>
        Deploy the same application into many different clusters even across
        cloud and hybrid environments, allowing change deployments across the fleet
        via Git and Cluster API.
      </>
    ),
  },
  {
    title: 'Trusted Application Delivery',
    icon: faHandshake,
    description: (
      <>
        Add policy as code to GitOps pipelines and enforce security and compliance,
        application resilience and coding standards from source to production.
        Validate policy conformance at every step in the software delivery pipeline:
        commit, build, deploy and run time.
      </>
    ),
  },
  {
    title: 'Progressive Delivery',
    icon: faBarsProgress,
    description: (
      <>
        Deploy into production environments safely using canary, blue/green and A/B
        strategies. Simple single file configuration defines success / rollback SLO
        using observability metrics from Prometheus, Datadog, New Relic and others.
      </>
    ),
  },
  {
    title: 'Self-Service Templates and Profiles',
    icon: faHandPointer,
    description: (
      <>
        Component profiles enable teams to deploy standard services quickly,
        consistently and reliably. Teams can curate the profiles that are available
        within their estate ensuring there is consistency everywhere. Using GitOps
        it's easy to guarantee the latest, secure versions of any component are
        deployed in all production systems.
      </>
    ),
  },
  {
    title: 'Health Status and Compliance Dashboards',
    icon: faHeartPulse,
    description: (
      <>
        Gain a single view of the health and state of the cluster and its workloads.
        Monitor deployments and alert on policy violations across apps and clusters.
      </>
    ),
  },
  {
    title: 'Team Workspaces',
    icon: faUsers,
    description: (
      <>
        Allow DevOps teams to work seamlessly together with multi-tenancy.
        Total RBAC control and policy enforcement with integration to enterprise IAM.
      </>
    ),
  },
  {
    title: 'Kubernetes Anywhere',
    icon: faMapLocationDot,
    description: (
      <>
        Reduce complexity with GitOps and install across all major target environments
        including support for on-premise, edge, hybrid, and multi-cloud Kubernetes clusters
      </>
    ),
  },
  {
    title: 'Critical 24/7 Support',
    icon: faBell,
    description: (
      <>
        Your business and workloads operate around the clock and so do we.
        Our team operates 24/7 so whenever there’s a problem our experts are
        there to help. Operate with confidence - we’ve got your back!
      </>
    ),
  },
  {
    title: '(CD) Pipelines',
    icon: faInfinity,
    description: (
      <>
        Rollout new software from development to production.
        Environment rollouts that work with your existing CI system.
      </>
    ),
  },
];

function Feature({icon, title, description, col}) {
  return (
    <div className={clsx('col', col)}>
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

export default function EEFeatures({col}) {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {EEFeatureList.map((props, idx) => (
            <Feature key={idx} {...props} col={col} />
          ))}
        </div>
      </div>
    </section>
  );
}
