import {Widget} from '@lumino/widgets';
import {FileBrowserModel} from '@jupyterlab/filebrowser';

export class Paging extends Widget {
  constructor(options: Paging.IOptions) {
    super();
    this._browser_model = options.browser_model;
    this._page_model = options.page_model;
    this.addClass('jp-DirPagination');
    this.update();
  }

  /**
   * Update the paging buttons
   */
  public update(): void {
    const max_page = this._page_model.max_page;
    const page = this._page_model.page;

    // Delete all page buttons.
    while (this.node.firstChild) {
      this.node.removeChild(this.node.firstChild);
    }

    // Don't show the paging widget if paging is not needed
    if (max_page === 1) {
      return;
    }

    /*
    if max_page <= 7
      show all items
    otherwise
    if page <= 4
     show 1-5 and the right ellipse
    if max_page-3 <= page
     show the left ellipse and max_page-4 to max_page
    if 4 < page < max_page-3
     show left ellipse, page-1, page, page+1, right ellipse
    */
    if (max_page <= 7) {
      for (let i = 1; i <= max_page; i++) {
        this._createButton(i, page);
      }
    } else if (page <= 4) {
      for (let i = 1; i <= 5; i++) {
        this._createButton(i, page);
      }
      this._createEllipse();
      this._createButton(max_page, page);
    } else if (page >= max_page - 3) {
      this._createButton(1, page);
      this._createEllipse();
      for (let i = max_page - 4; i <= max_page; i++) {
        this._createButton(i, page);
      }
    } else {
      this._createButton(1, page);
      this._createEllipse();
      this._createButton(page - 1, page);
      this._createButton(page, page);
      this._createButton(page + 1, page);
      this._createEllipse();
      this._createButton(max_page, page);
    }
  }

  private _createButton(i: number, activePage: number) {
    const button = document.createElement('a');
    button.textContent = i.toString();
    button.onclick = async () => {
      this._page_model.page = i;
      await this._browser_model.cd();
      this.update();
    };
    if (i === activePage) {
      button.className = 'active';
    }
    this.node.appendChild(button);
  }

  private _createEllipse() {
    const ellipse = document.createElement('span');
    ellipse.textContent = '...';
    this.node.appendChild(ellipse);
  }

  private _browser_model: FileBrowserModel;
  private _page_model: Paging.IModel;
}

export namespace Paging {
  /**
   * An options object for initializing a file browser directory listing.
   */
  export interface IOptions {
    /**
     * The file browser model instance.
     */
    browser_model: FileBrowserModel;
    /**
     * The pagination model instance.
     * This model is shared with the MountDrive class.
     */
    page_model: IModel;
  }

  export interface IModel {
    page: number;
    max_page: number;
  }
}
