import {TabPanel, Widget} from '@lumino/widgets';

import CreateExplore from './exploreView';
import CreatePipelineView from './pipelineView';
import CreateDatumView from './datumView';

export default function CreateTabs():Widget {
    var tabPanel = new TabPanel()
    tabPanel.addWidget(CreateExplore())//TODO Make CreateExplore not a method?
    tabPanel.addWidget(CreateDatumView())

    tabPanel.addWidget(CreatePipelineView());
    return tabPanel;

}