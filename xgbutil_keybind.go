/*
    All types and functions related to the keybind package.
    They live here because they rely on state in XUtil.
*/
package xgbutil

import "code.google.com/p/jamslam-x-go-binding/xgb"

// KeyBindCallback operates in the spirit of Callback, except that it works
// specifically on key bindings.
type KeyBindCallback interface {
    Connect(xu *XUtil, win xgb.Id, keyStr string)
    Run(xu *XUtil, ev interface{})
}

// KeyBindKey is the type of the key in the map of keybindings.
// It essentially represents the tuple
// (event type, window id, modifier, keycode).
type KeyBindKey struct {
    Evtype int
    Win xgb.Id
    Mod uint16
    Code byte
}

type KeyboardMapping struct {
    *xgb.GetKeyboardMappingReply
}

type ModifierMapping struct {
    *xgb.GetModifierMappingReply
}

// AttackKeyBindCallback associates an (event, window, mods, keycode)
// with a callback.
func (xu *XUtil) AttachKeyBindCallback(evtype int, win xgb.Id,
                                       mods uint16, keycode byte,
                                       fun KeyBindCallback) {
    // Create key
    key := KeyBindKey{evtype, win, mods, keycode}

    // Do we need to allocate?
    if _, ok := xu.keybinds[key]; !ok {
        xu.keybinds[key] = make([]KeyBindCallback, 0)
    }

    xu.keybinds[key] = append(xu.keybinds[key], fun)
    xu.keygrabs[key] += 1
}

// KeyBindKeys returns a copy of all the keys in the 'keybinds' map.
func (xu *XUtil) KeyBindKeys() []KeyBindKey {
    keys := make([]KeyBindKey, len(xu.keybinds))
    i := 0
    for key, _ := range xu.keybinds {
        keys[i] = key
        i++
    }
    return keys
}

// UpdateKeyBindKey takes a key bind key and a new key code.
// It will then remove the old key from keybinds and keygrabs,
// and add the new key with the old key's data into keybinds and keygrabs.
func (xu *XUtil) UpdateKeyBindKey(key KeyBindKey, newKc byte) {
    newKey := KeyBindKey{key.Evtype, key.Win, key.Mod, newKc}

    // Save old info
    oldCallbacks := xu.keybinds[key]
    oldGrabs := xu.keygrabs[key]

    // Delete old keys
    delete(xu.keybinds, key)
    delete(xu.keygrabs, key)

    // Add new keys with old info
    xu.keybinds[newKey] = oldCallbacks
    xu.keygrabs[newKey] = oldGrabs
}

// RunKeyBindCallbacks executes every callback corresponding to a
// particular event/window/mod/key tuple.
func (xu *XUtil) RunKeyBindCallbacks(event interface{}, evtype int,
                                     win xgb.Id, mods uint16, keycode byte) {
    // Create key
    key := KeyBindKey{evtype, win, mods, keycode}

    for _, cb := range xu.keybinds[key] {
        cb.Run(xu, event)
    }
}

// ConnectedKeyBind checks to see if there are any key binds for a particular
// event type already in play. This is to work around comparing function
// pointers (not allowed in Go), which would be used in 'Connected'.
func (xu *XUtil) ConnectedKeyBind(evtype int, win xgb.Id) bool {
    // Since we can't create a full key, loop through all key binds
    // and check if evtype and window match.
    for key, _ := range xu.keybinds {
        if key.Evtype == evtype && key.Win == win {
            return true
        }
    }

    return false
}

// DetachKeyBindWindow removes all callbacks associated with a particular
// window and event type (either KeyPress or KeyRelease)
// Also decrements the counter in the corresponding 'keygrabs' map
// appropriately.
func (xu *XUtil) DetachKeyBindWindow(evtype int, win xgb.Id) {
    // Since we can't create a full key, loop through all key binds
    // and check if evtype and window match.
    for key, _ := range xu.keybinds {
        if key.Evtype == evtype && key.Win == win {
            xu.keygrabs[key] -= len(xu.keybinds[key])
            delete(xu.keybinds, key)
        }
    }
}

// KeyBindGrabs returns the number of grabs on a particular
// event/window/mods/keycode combination. Namely, this combination
// uniquely identifies a grab. If it's repeated, we get BadAccess.
func (xu *XUtil) KeyBindGrabs(evtype int, win xgb.Id, mods uint16,
                              keycode byte) int {
    key := KeyBindKey{evtype, win, mods, keycode}
    return xu.keygrabs[key] // returns 0 if key does not exist
}

// KeyMapGet accessor
func (xu *XUtil) KeyMapGet() *KeyboardMapping {
    return xu.keymap
}

// KeyMapSet simply updates XUtil.keymap
func (xu *XUtil) KeyMapSet(keyMapReply *xgb.GetKeyboardMappingReply) {
    xu.keymap = &KeyboardMapping{keyMapReply}
}

// ModMapGet accessor
func (xu *XUtil) ModMapGet() *ModifierMapping {
    return xu.modmap
}

// ModMapSet simply updates XUtil.modmap
func (xu *XUtil) ModMapSet(modMapReply *xgb.GetModifierMappingReply) {
    xu.modmap = &ModifierMapping{modMapReply}
}
