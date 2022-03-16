import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'Easy to Use',
    // Svg: require('../../static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        Weave GitOps is designed to be effective and easy to get started with, 
        without compromising on power and capabilities.
      </>
    ),
  },
  {
    title: 'Focus on What Matters',
    // Svg: require('../../static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        Weave GitOps takes care of continuous deployment, Helm, Kustomization,
        integration with GitHub and Gitlab, so you can focus on coding.
      </>
    ),
  },
  {
    title: 'Powered by Flux',
    // Svg: require('../../static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        <a href="https://fluxcd.io">Flux</a> is a leading CNCF project around GitOps automation.
        Weave GitOps builds on this foundation to create a highly effective GitOps runtime.
      </>
    ),
  },
];

// function Feature({Svg, title, description}) {
//   return (
//     <div className={clsx('col col--4')}>
//       <div className="text--center">
//         <Svg className={styles.featureSvg} alt={title} />
//       </div>
//       <div className="text--center padding-horiz--md">
//         <h3>{title}</h3>
//         <p>{description}</p>
//       </div>
//     </div>
//   );
// }

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}


export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
