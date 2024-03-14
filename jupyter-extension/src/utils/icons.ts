import {LabIcon} from '@jupyterlab/ui-components';

// This cannot be imported via the module system because our webpack config imports svg wrapped as ReactComponent.
// LabIcon expects a string in order to properly validate the content.
// `class="jp-icon3"` must be added to the first path element in this svg to be a LabIcon
const fileSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg viewBox="0 0 21 31" width="21px" height="31px"
  xmlns="http://www.w3.org/2000/svg">
  <path class="jp-icon3" d="M.85.001C.365.055-.003.491 0 1.008v28.19c0 .556.425 1.007.95 1.007h18.98c.524 0 .949-.45.949-1.007V6.608a1.04 1.04 0 00-.277-.713l-5.299-5.6a.923.923 0 00-.662-.294H.949a.895.895 0 00-.099 0zm1.048 2.014h11.705v4.698c0 .556.425 1.007.949 1.007h4.429v20.472H1.898V2.015zm13.603 1.332l2.234 2.36h-2.234v-2.36zM4.646 12.083a.949.949 0 00-.797.549c-.154.32-.135.704.05 1.007a.935.935 0 00.846.457h6.327a.94.94 0 00.833-.499 1.06 1.06 0 000-1.015.94.94 0 00-.833-.5H4.745a.895.895 0 00-.099 0zm0 4.698a.949.949 0 00-.797.549c-.154.32-.135.705.05 1.007a.935.935 0 00.846.458h11.389a.94.94 0 00.833-.5 1.06 1.06 0 000-1.015.94.94 0 00-.833-.499H4.745a.895.895 0 00-.099 0zm0 4.699a.949.949 0 00-.797.548c-.154.321-.135.705.05 1.007a.935.935 0 00.846.458h11.389a.94.94 0 00.833-.5 1.06 1.06 0 000-1.014.94.94 0 00-.833-.5H4.745a.895.895 0 00-.099 0z" fill="#582F6B" fill-rule="nonzero"/>
</svg>`;

// This cannot be imported via the module system because our webpack config imports svg wrapped as ReactComponent.
// LabIcon expects a string in order to properly validate the content.
// `class="jp-icon3"` must be added to the first path element in this svg to be a LabIcon
const mountLogoSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="20px" height="20px" viewBox="0 0 20 20" version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink">
    <defs>
        <path class="jp-icon3" d="M14,13 L14,17.041 L10,19 L6,17.041 L6,13 L10,14.8365 L14,13 Z M10,10 L14,12.0852339 L10,14 L6,12.0852339 L10,10 Z M10.9899912,1 L10.9899912,5.72256 L12.5999411,4.16117749 L14,5.51882251 L10.7000295,8.71882251 C10.3617408,9.04686292 9.83953975,9.08786797 9.455273,8.84183766 L9.37515866,8.78472348 L9.29997054,8.71882251 L6,5.51882251 L7.40005893,4.16117749 L9.01000884,5.72256 L9.01000884,1 L10.9899912,1 Z" id="path-1"></path>
    </defs>
    <g id="Icons" stroke="none" stroke-width="1" fill-rule="evenodd">
        <g id="Icon/notebook-mount/black">
            <rect id="Bounding-Box" class="jp-icon3" fill-opacity="0" fill="#FFFFFF" x="0" y="0" width="20" height="20"></rect>
            <mask id="mask-2" fill="white">
                <use xlink:href="#path-1"></use>
            </mask>
            <use id="Combined-Shape" fill="#926AA1" fill-rule="nonzero" xlink:href="#path-1"></use>
            <g class="jp-icon3" id="Group" mask="url(#mask-2)" fill="#0F1012">
                <g id="Color">
                    <rect x="0" y="0" width="20" height="20"></rect>
                </g>
            </g>
        </g>
    </g>
</svg>
`;

// This cannot be imported via the module system because our webpack config imports svg wrapped as ReactComponent.
// LabIcon expects a string in order to properly validate the content.
// `class="jp-icon3"` must be added to the first path element in this svg to be a LabIcon
const repoSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="20px" height="20px" viewBox="0 0 20 20" version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink">
    <title>Icon/notebook-repo/black</title>
    <defs>
        <path class="jp-icon3" d="M18,7 L18,15.082 L10,19 L2,15.082 L2,7 L10,10.673 L18,7 Z M10,1 L18,5.17046787 L10,9 L2,5.17046787 L10,1 Z" id="path-1"></path>
    </defs>
    <g id="Icons" stroke="none" stroke-width="1" fill-rule="evenodd">
        <g id="Icon/notebook-repo/black">
            <mask id="mask-2" fill="white">
                <use xlink:href="#path-1"></use>
            </mask>
            <use id="Combined-Shape" fill="#926AA1" fill-rule="nonzero" xlink:href="#path-1"></use>
            <g id="Group" mask="url(#mask-2)" fill="#0F1012">
                <g id="Color">
                    <rect x="0" y="0" width="20" height="20"></rect>
                </g>
            </g>
        </g>
    </g>
</svg>`;

// This cannot be imported via the module system because our webpack config imports svg wrapped as ReactComponent.
// LabIcon expects a string in order to properly validate the content.
// `class="jp-icon3"` must be added to the first path element in this svg to be a LabIcon
const infoSvg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="20" height="20"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink">
    <defs>
        <path d="M10 0c5.5228475 0 10 4.4771525 10 10s-4.4771525 10-10 10S0 15.5228475 0 10 4.4771525 0 10 0Zm0 2c-4.418278 0-8 3.581722-8 8s3.581722 8 8 8 8-3.581722 8-8-3.581722-8-8-8Zm1.52 7.76v6.672H8.48V9.76h3.04ZM10 3.568c.5546667 0 1.0026667.14666667 1.344.44.3413333.29333333.512.66933333.512 1.128 0 .48-.1706667.87733333-.512 1.192-.3413333.31466667-.7893333.472-1.344.472-.55466667 0-1.00266667-.15466667-1.344-.464-.34133333-.30933333-.512-.69333333-.512-1.152 0-.45866667.17066667-.84266667.512-1.152.34133333-.30933333.78933333-.464 1.344-.464Z" id="infoa"/>
    </defs>
    <g fill-rule="evenodd">
        <path d="M0 0h20v20H0z"/>
        <mask id="infob" fill="#fff">
            <use xlink:href="#infoa"/>
        </mask>
        <use fill="#000" fill-rule="nonzero" xlink:href="#infoa"/>
        <g mask="url(#infob)" fill="currentcolor">
            <path class="jp-icon3" d="M0 0h20v20H0z"/>
        </g>
    </g>
</svg>`;

export const fileIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:file',
  svgstr: fileSvg,
});

export const mountLogoIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:mount-logo',
  svgstr: mountLogoSvg,
});

export const repoIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:repo',
  svgstr: repoSvg,
});

export const infoIcon = new LabIcon({
  name: 'jupyterlab-pachyderm:info',
  svgstr: infoSvg,
});
