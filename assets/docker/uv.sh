
SAMPLE_UV_VER="0.9.1"
install() {
    UV_VER=$1
    SCRIPT="uv-installer.sh"
    echo "eeee yooo ${UV_VER}"
    wget https://github.com/astral-sh/uv/releases/download/${UV_VER}/${SCRIPT}
    chmod +x ${SCRIPT}
    UV_INSTALL_DIR=/usr/local/bin UV_NO_MODIFY_PATH=1 sh ${SCRIPT}
}

    # command: /bin/bash -lc "source watcher.sh"
SAMPLE_PY_VER="3.11"
spawn_venv() {
    PY_VER=$1
    GLOBVENV="/globvenv"
    uv venv $GLOBVENV --python $PY_VER
    # echo source $GLOBVENV/bin/activate >> /root/.bashrc
    PROFILE_SH=/etc/profile.d/uv.sh
    echo . $GLOBVENV/bin/activate >> $PROFILE_SH
    chmod +x $PROFILE_SH
}