# TODO for testing, remove later
def start_debugger():
    return
    import pydevd_pycharm as pydevd
    pydevd.settrace(host="localhost", port=32000, suspend=False, stdoutToServer=True, stderrToServer=True)