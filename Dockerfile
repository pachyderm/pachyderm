FROM jupyter/minimal-notebook:177037d09156

RUN pip install --no-cache-dir astropy
RUN pip install ipyauth
RUN jupyter serverextension enable --py --sys-prefix ipyauth.ipyauth_callback
RUN jupyter nbextension enable --py --sys-prefix ipyauth.ipyauth_widget 
